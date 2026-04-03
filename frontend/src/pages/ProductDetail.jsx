import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import api from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import { ChevronLeft, Plus, Minus, ShoppingCart, Loader2, Star, CheckCircle2 } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

const ProductDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const addItem = useCartStore(state => state.addItem);
  
  const [product, setProduct] = useState(null);
  const [relatedProducts, setRelatedProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [quantity, setQuantity] = useState(1);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    window.scrollTo(0, 0);
    fetchProduct();
  }, [id]);

  const fetchProduct = async () => {
    setLoading(true);
    try {
      const [prodRes, allProdsRes, catRes] = await Promise.all([
        api.get(`/catalog/products`), // Backend doesn't have direct GetProductByID? Let's check.
        api.get('/catalog/products'),
        api.get('/catalog/categories')
      ]);

      const allProds = Array.isArray(allProdsRes.data) ? allProdsRes.data : [];
      const foundProduct = allProds.find(p => p.id === parseInt(id));
      
      if (!foundProduct) {
        navigate('/');
        return;
      }

      setProduct(foundProduct);
      setCategories(catRes.data || []);
      
      // Filter related (same category, different ID)
      const related = allProds
        .filter(p => p.category_id === foundProduct.category_id && p.id !== foundProduct.id)
        .slice(0, 4);
      setRelatedProducts(related);

    } catch (err) {
      console.error(err);
      navigate('/');
    } finally {
      setLoading(false);
    }
  };

  const handleAddToCart = () => {
    if (!product) return;
    for (let i = 0; i < quantity; i++) {
      addItem(product);
    }
    setAdded(true);
    setTimeout(() => setAdded(false), 2000);
  };

  const getImageUrl = (url) => {
    if (!url) return null;
    return url;
  };

  if (loading) return (
    <div className="loading-screen">
      <div className="spinner" />
      <p>Yuklanmoqda...</p>
    </div>
  );

  const categoryName = categories.find(c => c.id === product?.category_id)?.name || 'Kategoriya';

  return (
    <div className="pd-page animate-fade">
      {/* Header / Back */}
      <div className="pd-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          <ChevronLeft size={24} />
        </button>
        <span className="header-title">Mahsulot tafsilotlari</span>
        <div style={{ width: 40 }} /> {/* Spacer */}
      </div>

      <div className="pd-container">
        {/* Left: Image */}
        <div className="pd-image-section">
          <motion.div 
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            className="pd-main-img-wrap"
          >
            {getImageUrl(product.image_url) ? (
              <img src={getImageUrl(product.image_url)} alt={product.name} className="pd-main-img" />
            ) : (
              <div className="pd-img-placeholder">🍽</div>
            )}
          </motion.div>
        </div>

        {/* Right: Info */}
        <div className="pd-info-section">
          <div className="pd-badge">{categoryName}</div>
          <h1 className="pd-title">{product.name}</h1>
          
          <div className="pd-rating">
            <div className="stars-mini">
              {[1, 2, 3, 4, 5].map(s => <Star key={s} size={14} fill="#fbbf24" color="#fbbf24" />)}
            </div>
            <span className="rating-count">4.9 (120+ baholar)</span>
          </div>

          <div className="pd-price-row">
            <span className="pd-price">{(product.price).toLocaleString()} so'm</span>
            <div className="qty-control">
              <button disabled={quantity <= 1} onClick={() => setQuantity(q => q - 1)}><Minus size={18} /></button>
              <span>{quantity}</span>
              <button onClick={() => setQuantity(q => q + 1)}><Plus size={18} /></button>
            </div>
          </div>

          <div className="pd-description">
            <h3>Tavsif</h3>
            <p>{product.description || "Ushbu mahsulot uchun hozircha tavsif qo'shilmagan. Lekin u juda mazali ekanligiga shubha yo'q!"}</p>
          </div>

          <div className="pd-features">
            <div className="feature-item">
              <CheckCircle2 size={18} color="var(--success)" />
              <span>Yangi mahsulotlar</span>
            </div>
            <div className="feature-item">
              <CheckCircle2 size={18} color="var(--success)" />
              <span>Tez yetkazib berish</span>
            </div>
          </div>

          <button 
            className={`pd-add-btn ${added ? 'success' : ''}`} 
            onClick={handleAddToCart}
            disabled={added}
          >
            {added ? (
              <><CheckCircle2 size={20} /> Savatga qo'shildi</>
            ) : (
              <><ShoppingCart size={20} /> Savatga qo'shish</>
            )}
          </button>
        </div>
      </div>

      {/* Related Products */}
      {relatedProducts.length > 0 && (
        <section className="related-section">
          <div className="section-header">
            <div className="section-pill">O'xshash</div>
            <h2>Boshqa mahsulotlar</h2>
          </div>
          
          <div className="related-grid">
            {relatedProducts.map(item => (
              <Link key={item.id} to={`/product/${item.id}`} className="related-card">
                <div className="rel-img-wrap">
                  {item.image_url ? (
                    <img src={getImageUrl(item.image_url)} alt={item.name} />
                  ) : (
                    <div className="rel-placeholder">🍽</div>
                  )}
                </div>
                <div className="rel-body">
                  <h4>{item.name}</h4>
                  <p>{item.price.toLocaleString()} so'm</p>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      <style>{`
        .pd-page {
          padding-bottom: 5rem;
        }

        .pd-header {
          display: flex; align-items: center; justify-content: space-between;
          padding: 1rem; position: sticky; top: 0; z-index: 100;
          background: rgba(13,13,15,0.8); backdrop-filter: blur(20px);
        }

        .back-btn {
          width: 40px; height: 40px; border-radius: 50%;
          background: var(--bg-surface); color: white; display: flex; align-items: center; justify-content: center;
          border: 1px solid var(--border); transition: 0.2s;
        }
        .back-btn:hover { border-color: var(--primary); color: var(--primary); }

        .header-title { font-weight: 700; font-size: 0.95rem; text-transform: uppercase; letter-spacing: 0.05em; }

        .pd-container {
          display: grid; grid-template-columns: 1.1fr 0.9fr; gap: 3rem;
          padding: 2rem; max-width: 1200px; margin: 0 auto;
        }

        .pd-main-img-wrap {
          aspect-ratio: 1; border-radius: 24px; overflow: hidden;
          background: var(--bg-card); position: sticky; top: 100px;
          border: 1px solid var(--border); box-shadow: 0 20px 40px rgba(0,0,0,0.3);
        }
        .pd-main-img { width: 100%; height: 100%; object-fit: cover; transition: 0.5s transform cubic-bezier(0.4, 0, 0.2, 1); }
        .pd-main-img-wrap:hover .pd-main-img { transform: scale(1.05); }

        .pd-img-placeholder { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; font-size: 8rem; opacity: 0.1; }

        .pd-info-section { display: flex; flex-direction: column; gap: 1rem; }
        
        .pd-badge {
          display: inline-flex; width: fit-content; padding: 0.4rem 1rem;
          background: rgba(249,115,22,0.12); border: 1px solid rgba(249,115,22,0.25);
          color: var(--primary); border-radius: 99px; font-size: 0.75rem; font-weight: 800; text-transform: uppercase;
        }

        .pd-title { font-size: 2.8rem; letter-spacing: -0.02em; line-height: 1.1; }

        .pd-rating { display: flex; align-items: center; gap: 0.75rem; }
        .stars-mini { display: flex; gap: 2px; }
        .rating-count { color: var(--text-muted); font-size: 0.85rem; }

        .pd-price-row {
          display: flex; align-items: center; justify-content: space-between;
          padding: 1.5rem 2rem; background: var(--bg-card); border-radius: 20px;
          border: 1px solid var(--border); margin: 1rem 0;
        }
        .pd-price { font-size: 2rem; font-weight: 800; color: var(--primary-light); }

        .qty-control {
          display: flex; align-items: center; gap: 1.25rem;
          background: var(--bg-base); padding: 0.5rem 0.75rem; border-radius: 12px; border: 1px solid var(--border);
        }
        .qty-control button {
          width: 34px; height: 34px; display: flex; align-items: center; justify-content: center;
          background: var(--bg-surface); color: white; border-radius: 8px; transition: 0.2s;
        }
        .qty-control button:hover:not(:disabled) { background: var(--primary); }
        .qty-control button:disabled { opacity: 0.3; cursor: not-allowed; }
        .qty-control span { font-weight: 800; font-size: 1.2rem; min-width: 24px; text-align: center; }

        .pd-description h3 { font-size: 1.1rem; margin-bottom: 0.5rem; }
        .pd-description p { color: var(--text-secondary); line-height: 1.6; font-size: 0.95rem; }

        .pd-features { display: flex; flex-direction: column; gap: 0.6rem; margin: 0.5rem 0; }
        .feature-item { display: flex; align-items: center; gap: 0.6rem; font-size: 0.85rem; color: var(--text-secondary); }

        .pd-add-btn {
          margin-top: 1rem; width: 100%; padding: 1.25rem; border-radius: 18px;
          background: var(--grad-brand); color: white; font-weight: 800; font-size: 1.1rem;
          display: flex; align-items: center; justify-content: center; gap: 0.75rem;
          box-shadow: 0 10px 30px rgba(249,115,22,0.4); transition: 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
        }
        .pd-add-btn:hover:not(:disabled) { transform: translateY(-3px); box-shadow: 0 15px 40px rgba(249,115,22,0.55); }
        .pd-add-btn:active { transform: scale(0.97); }
        .pd-add-btn.success { background: var(--success); box-shadow: 0 10px 30px rgba(16,185,129,0.3); }

        /* RELATED SECTION */
        .related-section { padding: 4rem 2rem; max-width: 1200px; margin: 0 auto; border-top: 1px solid var(--border); }
        .related-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1.5rem; margin-top: 2rem; }
        
        .related-card {
          background: var(--bg-card); border: 1px solid var(--border); border-radius: 20px; overflow: hidden;
          transition: 0.3s; text-decoration: none;
        }
        .related-card:hover { transform: translateY(-5px); border-color: var(--primary); box-shadow: var(--shadow-md); }
        .rel-img-wrap { width: 100%; height: 160px; background: rgba(255,255,255,0.03); }
        .rel-img-wrap img { width: 100%; height: 100%; object-fit: cover; }
        .rel-placeholder { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; font-size: 3rem; opacity: 0.1; }
        
        .rel-body { padding: 1rem; text-align: center; }
        .rel-body h4 { font-size: 0.95rem; margin-bottom: 0.25rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
        .rel-body p { font-weight: 800; color: var(--primary-light); font-size: 0.9rem; }

        @media (max-width: 900px) {
          .pd-container { grid-template-columns: 1fr; gap: 2rem; padding: 1.25rem; }
          .pd-title { font-size: 2.2rem; }
          .pd-image-section { order: -1; }
          .pd-main-img-wrap { position: static; max-width: 500px; margin: 0 auto; }
          .related-grid { grid-template-columns: repeat(2, 1fr); gap: 1rem; }
        }

        @media (max-width: 480px) {
          .pd-title { font-size: 1.8rem; }
          .pd-price-row { padding: 1rem 1.25rem; }
          .pd-price { font-size: 1.5rem; }
          .pd-add-btn { padding: 1rem; font-size: 1rem; }
          .related-grid { grid-template-columns: 1fr; }
          .pd-main-img-wrap { max-width: 100%; }
        }
      `}</style>
    </div>
  );
};

export default ProductDetail;
